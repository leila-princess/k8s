package service

import (
	"archive/tar"
	"archive/zip"
	"backend/app/model"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gorm.io/gorm"
)

// ===== 入口：从公共镜像 + 代码压缩包 构建并推送到租户项目 =====
//
// BaseProject: 你公共基础镜像所在的项目名（例如 "base"）
// BaseIamgeName:    基础镜像仓库名（例如 "alpine" 或 "python"）
// BaseTag:     基础镜像标签（例如 "3.18" / "3.10"）
// ArchivePath: 代码压缩包本地路径（支持 .tar.gz/.tgz/.tar/.zip）
// TenantName:  要推送到的租户（项目同名）
// TargetIamgeName:  目标仓库名（例如 "myapp"）
// TargetTag:   目标标签（例如 "v1.0.0"）
func (s *HarborService) BuildAndPushFromArchive(
	ctx context.Context,
	BaseProject, BaseIamgeName, BaseTag string,
	ArchivePath string,
	TenantName string,
	TargetIamgeName string,
	TargetTag string,
) error {
	adminCfg := NewHarborConfig()
	tenantCfg := NewTenantHarborConfig(TenantName)

	// 1) 确保租户项目存在（admin 创建最稳；你只要求“拉镜像”不用 admin，这里仍可保留）
	if err := s.ensureProject(ctx, TenantName); err != nil {
		return fmt.Errorf("ensure project %q failed: %w", TenantName, err)
	}

	// 2) 使用“租户账号”拉基础镜像（不使用 admin）
	//    repository 传 "<project>/<repo>"，tag 传 "<tag>"
	baseRepository := fmt.Sprintf("%s/%s", BaseProject, BaseIamgeName)
	if err := s.PullImageWithAuth(TenantName, baseRepository, BaseTag); err != nil {
		return fmt.Errorf("pull base image with tenant auth %s:%s failed: %w",
			baseRepository, BaseTag, err)
	}

	// 3) 准备构建上下文：Dockerfile 的 FROM 需要完整的 baseRef（含 registry）
	//    这里用 adminCfg.DockerHost 仅作为“镜像地址字符串”，实际 Build 时会提供 auths（含租户）
	baseRef := fmt.Sprintf("%s/%s/%s:%s", adminCfg.DockerHost, BaseProject, BaseIamgeName, BaseTag)
	buildCtxTar, cleanup, err := s.prepareBuildContextTar(baseRef, ArchivePath)
	if err != nil {
		return fmt.Errorf("prepare build context failed: %w", err)
	}
	defer cleanup()

	// 4) docker build —— 只打“本地标签”，例如 "myapp:v1.0.0"
	localTag := fmt.Sprintf("%s:%s", TargetIamgeName, TargetTag)
	if err := s.dockerBuild(ctx, buildCtxTar, localTag, adminCfg, tenantCfg); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	// 5) 推送镜像到 <tenant>/<repo>:<tag> —— 复用已有 PushImage（会负责 retag + push）
	if err := s.PushImage(TenantName, TargetIamgeName, TargetTag); err != nil {
		return fmt.Errorf("push image failed: %w", err)
	}

	return nil
}

// ===== 项目存在性保证 =====

func (s *HarborService) ensureProject(ctx context.Context, project string) error {
	admin := NewHarborConfig()
	// 先查
	getURL := fmt.Sprintf("%s/api/v2.0/projects/%s", admin.Host, project)
	{
		req, _ := http.NewRequestWithContext(ctx, "GET", getURL, nil)
		req.SetBasicAuth(admin.Username, admin.Password)
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil // 已存在
			}
		}
	}

	// 不存在则创建
	createURL := fmt.Sprintf("%s/api/v2.0/projects", admin.Host)
	payload := strings.NewReader(fmt.Sprintf(`{"project_name":"%s","metadata":{"public":"false"}}`, project))
	req, _ := http.NewRequestWithContext(ctx, "POST", createURL, payload)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(admin.Username, admin.Password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		// 201: 创建成功；409: 已存在（并发/瞬时）
		return nil
	}
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create project %q failed: http %d: %s", project, resp.StatusCode, string(b))
}

// ===== 构建上下文（本地解压到 src/ + Dockerfile -> 打包成 tar.Reader） =====

func (s *HarborService) prepareBuildContextTar(baseRef, archivePath string) (io.ReadCloser, func(), error) {
	// 1) 创建临时目录
	tmpDir, err := os.MkdirTemp("", "buildctx-*")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		cleanup()
		return nil, nil, err
	}

	// 2) 解压到 src/
	if err := extractArchiveTo(archivePath, srcDir); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("extract %q failed: %w", archivePath, err)
	}

	// 3) 写 Dockerfile（FROM 使用完整 baseRef）
	df := `FROM %s
WORKDIR /app
COPY src/ /app/
# 你可以在这里添加你的构建步骤，比如：
# RUN pip install -r requirements.txt  或 RUN go build ./...
`
	dockerfile := fmt.Sprintf(df, baseRef)
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(dockerfile), 0o644); err != nil {
		cleanup()
		return nil, nil, err
	}

	// 4) 打包整个 tmpDir 为 tar（docker build 上下文）
	pr, pw := io.Pipe()
	go func() {
		err := tarDirectory(tmpDir, pw)
		_ = pw.CloseWithError(err)
	}()
	return pr, cleanup, nil
}

// ===== docker build =====
// 这里仅打“本地标签”，例如 "myapp:v1.0.0"
func (s *HarborService) dockerBuild(ctx context.Context, buildCtx io.Reader, localTag string, adminCfg, tenantCfg *HarborConfig) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	// 为了在 build 阶段能拉私有镜像，这里把 registry 的鉴权塞进去
	// 虽然我们已用租户账号预拉了基础镜像，但 Dockerfile 中的 FROM 在构建时仍可能访问 registry
	auths := map[string]types.AuthConfig{
		adminCfg.DockerHost: {
			Username:      adminCfg.Username,
			Password:      adminCfg.Password,
			ServerAddress: adminCfg.DockerHost,
		},
		tenantCfg.DockerHost: {
			Username:      tenantCfg.Username,
			Password:      tenantCfg.Password,
			ServerAddress: tenantCfg.DockerHost,
		},
	}

	resp, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Tags:           []string{localTag}, // 只打本地标签
		Dockerfile:     "Dockerfile",
		Remove:         true, // 构建完成清理中间层
		ForceRemove:    true,
		SuppressOutput: false,
		AuthConfigs:    auths,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return drainJSONStream(resp.Body, "BUILD")
}

// ====== 工具：解压、打包、日志 ======

func extractArchiveTo(archivePath, dst string) error {
	l := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(l, ".tar.gz") || strings.HasSuffix(l, ".tgz"):
		return untarGz(archivePath, dst)
	case strings.HasSuffix(l, ".tar"):
		return untar(archivePath, dst)
	case strings.HasSuffix(l, ".zip"):
		return unzip(archivePath, dst)
	default:
		return fmt.Errorf("unsupported archive format: %s (支持 .tar.gz/.tgz/.tar/.zip)", archivePath)
	}
}

func untarGz(p, dst string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	return untarStream(gzr, dst)
}

func untar(p, dst string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return untarStream(f, dst)
}

func untarStream(r io.Reader, dst string) error {
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dst, h.Name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, fs.FileMode(h.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(h.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return err
			}
			_ = out.Close()
		default:
			// 其它类型简单忽略（可按需扩展支持 Symlink/Hardlink 等）
		}
	}
}

func unzip(p, dst string) error {
	zr, err := zip.OpenReader(p)
	if err != nil {
		return err
	}
	defer zr.Close()
	for _, f := range zr.File {
		target := filepath.Join(dst, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
	}
	return nil
}

func tarDirectory(root string, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		// 修正 header.Name 为相对路径
		hdr.Name = rel

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()
		}
		return nil
	})
}

// 读取 docker 的 JSON 流式日志（build/pull/push），友好打印错误
func drainJSONStream(r io.Reader, phase string) error {
	type line struct {
		Stream string `json:"stream"`
		Status string `json:"status"`
		Error  string `json:"error"`
		// 其它字段按需扩展
	}
	dec := json.NewDecoder(bufio.NewReader(r))
	var any bool
	for {
		var l line
		if err := dec.Decode(&l); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// 有些 daemon 会返回普通文本流；兜底把剩余内容读完打印
			slurp, _ := io.ReadAll(dec.Buffered())
			if len(strings.TrimSpace(string(slurp))) > 0 {
				fmt.Printf("[%s] %s\n", phase, string(slurp))
			}
			// 非致命，直接退出
			break
		}
		any = true
		if l.Stream != "" {
			fmt.Print(l.Stream)
		}
		if l.Status != "" {
			fmt.Printf("[%s] %s\n", phase, l.Status)
		}
		if l.Error != "" {
			return fmt.Errorf("%s error: %s", phase, l.Error)
		}
	}
	// 如果完全不是 JSON（例如 http body 是纯文本），再兜底读一遍
	if !any {
		all, _ := io.ReadAll(r)
		txt := strings.TrimSpace(string(all))
		if strings.Contains(strings.ToLower(txt), "error") {
			return fmt.Errorf("%s error: %s", phase, txt)
		}
		if txt != "" {
			fmt.Printf("[%s] %s\n", phase, txt)
		}
	}
	return nil
}

func (s *HarborService) AsyncBuildAndPushFromArchive(
	db *gorm.DB,
	task *model.ImageTask,
	baseProject, BaseIamgeName, baseTag string,
	archivePath string,
) error {
	go func() {
		// 事务开始
		tx := db.Begin()
		if tx.Error != nil {
			fmt.Printf("[AsyncBuildAndPush] 开启事务失败: %v\n", tx.Error)
			return
		}

		// 置为运行中
		task.Status = model.TaskStatusRunning
		task.Message = "开始构建与推送"
		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			fmt.Printf("[AsyncBuildAndPush] 更新任务为运行中失败: %v\n", err)
			return
		}

		// 设置超时上下文（按需调整）
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancel()

		// 调用同步构建 + 推送
		err := s.BuildAndPushFromArchive(
			ctx,
			baseProject, BaseIamgeName, baseTag,
			archivePath,
			task.Username,  // TenantName（项目同名）
			task.ImageName, // TargetIamgeName
			task.Tag,       // TargetTag
		)

		if err != nil {
			// 回滚事务，然后用新的连接写入失败状态（与你现有风格一致）
			tx.Rollback()
			task.Status = model.TaskStatusFailed
			task.Message = fmt.Sprintf("构建/推送失败: %v", err)
			if e := db.Model(task).Updates(map[string]interface{}{
				"status":  task.Status,
				"message": task.Message,
			}).Error; e != nil {
				fmt.Printf("[AsyncBuildAndPush] 写入失败状态出错: %v\n", e)
			}
			return
		}

		task.Status = model.TaskStatusComplete
		task.Message = "构建并推送成功"
		if err := tx.Save(task).Error; err != nil {
			tx.Rollback()
			fmt.Printf("[AsyncBuildAndPush] 保存完成状态失败: %v\n", err)
			return
		}
		tx.Commit()
	}()
	return nil
}

// （可选）便捷同步包装：直接用 task 字段映射，不走异步。
func (s *HarborService) BuildAndPushByTask(
	ctx context.Context,
	task *model.ImageTask,
	baseProject, BaseIamgeName, baseTag string,
	archivePath string,
) error {
	return s.BuildAndPushFromArchive(
		ctx,
		baseProject, BaseIamgeName, baseTag,
		archivePath,
		task.Username,  // TenantName
		task.ImageName, // TargetIamgeName
		task.Tag,       // TargetTag
	)
}
