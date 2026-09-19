import fs from 'node:fs/promises';
import { Presentation, PresentationFile } from '@oai/artifact-tool';

const OUT = 'D:/ruoyi-go-master/kubernetes-ai-platform-architecture-interview-focus.pptx';
const PREVIEW = 'D:/ruoyi-go-master/.pptx_build/preview.png';
const LAYOUT = 'D:/ruoyi-go-master/.pptx_build/layout.json';
const FONT = 'Microsoft YaHei';

const C = {
  bg: '#F6F9FC', navy: '#0B4F93', cyan: '#2CB7D7', ink: '#17324A', text: '#42596B',
  blue: '#1479C9', blueSoft: '#D8EDF9', purpleSoft: '#E8E2F6', purpleLine: '#9D8BC2',
  greenSoft: '#E2F2EA', greenLine: '#65A985', orangeSoft: '#FFF0DC', orangeLine: '#E4872D',
  white: '#FFFFFF', darkLine: '#21364A', security: '#17324A'
};

async function writeBlob(path, blob) {
  await fs.writeFile(path, new Uint8Array(await blob.arrayBuffer()));
}

function setText(shape, text, opts = {}) {
  shape.text = text;
  shape.text.style = {
    typeface: FONT,
    fontSize: opts.size ?? 22,
    bold: opts.bold ?? false,
    color: opts.color ?? C.ink,
    alignment: opts.align ?? 'center',
    verticalAlignment: opts.valign ?? 'middle',
    autoFit: 'shrinkText',
    wrap: 'square',
    insets: opts.insets ?? { top: 4, right: 8, bottom: 4, left: 8 }
  };
}

function box(slide, name, x, y, w, h, fill, line, radius = 16, shadow = false) {
  return slide.shapes.add({
    geometry: 'roundRect', name,
    position: { left: x, top: y, width: w, height: h },
    fill, line: { style: 'solid', fill: line, width: 2 }, borderRadius: radius,
    shadow: shadow ? 'shadow-sm' : 'shadow-none'
  });
}

function textBox(slide, name, text, x, y, w, h, opts = {}) {
  const s = slide.shapes.add({
    geometry: 'textbox', name, position: { left: x, top: y, width: w, height: h },
    fill: 'none', line: { style: 'solid', fill: 'none', width: 0 }
  });
  setText(s, text, opts);
  return s;
}

function addModule(slide, name, x, y, w, h, title, lines, theme = 'purple') {
  const palette = theme === 'green'
    ? { fill: '#FFFFFF', line: C.greenLine }
    : theme === 'orange'
      ? { fill: '#FFF8EF', line: '#E3AE6D' }
      : { fill: '#F8F5FC', line: '#B5A6D4' };
  const s = box(slide, name, x, y, w, h, palette.fill, palette.line, 14, false);
  textBox(slide, `${name}-title`, title, x + 10, y + 12, w - 20, 31, { size: 23, bold: true });
  textBox(slide, `${name}-body`, lines.join('\n'), x + 10, y + 47, w - 20, h - 56, { size: 18, color: C.text });
  return s;
}

function connect(slide, from, to, opts = {}) {
  return slide.shapes.connect(from, to, {
    kind: opts.kind ?? 'straight', fromSide: opts.from ?? 'right', toSide: opts.to ?? 'left',
    line: { style: opts.dashed ? 'dashed' : 'solid', fill: opts.color ?? C.darkLine, width: opts.width ?? 3 },
    tail: { type: 'triangle', width: 'sm', length: 'sm' }
  });
}

async function main() {
  await fs.mkdir('D:/ruoyi-go-master/.pptx_build', { recursive: true });
  const deck = Presentation.create({ slideSize: { width: 1920, height: 1080 } });
  const slide = deck.slides.add();
  slide.background.fill = C.bg;

  // Header
  const header = slide.shapes.add({ geometry: 'rect', name: 'header', position: { left: 0, top: 0, width: 1920, height: 112 }, fill: C.navy, line: { style: 'solid', fill: C.navy, width: 0 } });
  textBox(slide, 'title', '基于 Kubernetes 的智能算力服务化平台', 72, 18, 1180, 58, { size: 50, bold: true, color: C.white, align: 'left' });
  textBox(slide, 'subtitle', '镜像标准化 · GitOps 自动交付 · 多租户资源隔离 · 统一流量治理 · 全栈可观测', 74, 76, 1100, 27, { size: 21, color: '#D8E9F8', align: 'left' });
  slide.shapes.add({ geometry: 'rect', name: 'cyan-divider', position: { left: 0, top: 105, width: 1920, height: 7 }, fill: C.cyan, line: { style: 'solid', fill: C.cyan, width: 0 } });
  const target = box(slide, 'target-state', 1720, 30, 140, 52, '#397CB6', '#397CB6', 10, false);
  setText(target, '目标态架构', { size: 20, bold: true, color: C.white });

  // Layer labels
  const layers = [
    ['访问层', 145, 102, C.blueSoft], ['平台控制层', 270, 245, C.purpleSoft],
    ['交付链路', 538, 139, C.greenSoft], ['运行与观测层', 700, 284, C.orangeSoft]
  ];
  for (const [label, y, h, fill] of layers) {
    const s = box(slide, `layer-${label}`, 36, y, 128, h, fill, fill, 16, false);
    setText(s, label, { size: 25, bold: true });
  }

  // Access layer nodes
  const user = box(slide, 'platform-user', 360, 145, 300, 102, C.white, '#7BAFD0', 16, true);
  setText(user, '平台用户 / 租户', { size: 23, bold: true });
  const gateway = box(slide, 'unified-entry', 940, 145, 420, 102, C.white, '#7BAFD0', 16, true);
  setText(gateway, '统一访问入口\nIngress Controller · HTTPS / 域名 / 路径路由', { size: 21, bold: true });

  // Control plane panel and modules
  const control = box(slide, 'go-control-plane', 205, 270, 1614, 245, C.white, C.purpleLine, 20, true);
  const controlHead = slide.shapes.add({ geometry: 'rect', name: 'control-plane-header', position: { left: 207, top: 272, width: 1610, height: 48 }, fill: C.purpleSoft, line: { style: 'solid', fill: C.purpleSoft, width: 0 } });
  textBox(slide, 'control-plane-title', 'Go 平台控制面（REST API + client-go）', 238, 276, 650, 38, { size: 26, bold: true, align: 'left' });
  const auth = addModule(slide, 'auth-tenant', 239, 343, 425, 134, '双层身份与租户授权', ['JWT + Redis → 租户身份', '租户 → Namespace + ServiceAccount', 'RoleBinding → Role → API Server RBAC 判定']);
  auth.fill = '#F4F0FB'; auth.line = { style: 'solid', fill: '#8C70BF', width: 2 };
  const build = addModule(slide, 'image-build', 692, 343, 535, 134, '模型环境标准化与镜像构建', ['基础镜像：Python / PyTorch / TensorFlow / CUDA', '代码压缩包 → src/ + 动态 Dockerfile → tar 流', 'Docker ImageBuild → 专属镜像 → 租户 Harbor 项目']);
  build.fill = '#EEF7FD'; build.line = { style: 'solid', fill: '#4B9BD3', width: 2 };
  const orchestrator = addModule(slide, 'k8s-orchestrator', 1255, 343, 530, 134, 'Kubernetes 编排与声明式交付', ['Deployment：镜像 / 启动命令 / 端口 / CPU / 内存 / GPU', 'Namespace · Quota · LimitRange · ServiceAccount / RBAC', 'Service / Ingress · Git 仓库 · Argo CD 同步']);
  orchestrator.fill = '#F1F8F4'; orchestrator.line = { style: 'solid', fill: '#65A985', width: 2 };

  // Delivery chain
  const source = addModule(slide, 'source-assets', 205, 538, 294, 139, '模型代码压缩包', ['.zip / .tar / .tar.gz / .tgz', '模型代码 · 依赖文件 · 启动脚本'], 'green');
  const pipeline = addModule(slide, 'docker-pipeline', 565, 538, 271, 139, 'Docker 构建上下文', ['tmpDir / Dockerfile + src/', '打包 tar.Reader → ImageBuild'], 'green');
  const harbor = addModule(slide, 'harbor', 902, 538, 238, 139, 'Harbor', ['私有镜像仓库', 'Tag / 版本 / 制品交付'], 'green');
  const configRepo = addModule(slide, 'git-config', 1206, 538, 242, 139, 'Git 配置仓库', ['租户化 YAML 清单', '唯一事实源 · 可审计回滚'], 'green');
  const argocd = addModule(slide, 'argocd-appset', 1514, 538, 305, 139, 'Argo CD ApplicationSet', ['按租户生成 Application', '自动同步 · 漂移修复 · 回滚'], 'green');

  // Runtime cluster and observability
  const cluster = box(slide, 'kubernetes-cluster', 205, 700, 1188, 284, C.white, C.orangeLine, 20, true);
  textBox(slide, 'cluster-title', 'Kubernetes 智能算力集群', 238, 714, 500, 38, { size: 27, bold: true, align: 'left' });
  const tenantA = box(slide, 'tenant-a', 239, 761, 548, 184, '#FFF8EF', '#E3AE6D', 16, false);
  const tenantB = box(slide, 'tenant-b', 811, 761, 548, 184, '#FFF8EF', '#E3AE6D', 16, false);
  const tenantAHead = box(slide, 'tenant-a-head', 255, 777, 516, 36, '#F7D8AB', '#F7D8AB', 9, false); setText(tenantAHead, 'Tenant A Namespace', { size: 21, bold: true });
  const tenantBHead = box(slide, 'tenant-b-head', 827, 777, 516, 36, '#F7D8AB', '#F7D8AB', 9, false); setText(tenantBHead, 'Tenant B Namespace', { size: 21, bold: true });
  textBox(slide, 'tenant-a-resources', 'ServiceAccount + RoleBinding\nResourceQuota + LimitRange\nDeployment / Pod', 267, 831, 275, 93, { size: 17, color: C.text, align: 'left' });
  textBox(slide, 'tenant-b-resources', 'ServiceAccount + RoleBinding\nResourceQuota + LimitRange\nDeployment / Pod', 839, 831, 275, 93, { size: 17, color: C.text, align: 'left' });
  const serviceA = box(slide, 'service-a', 551, 832, 196, 83, C.white, '#E3AE6D', 12, false); setText(serviceA, '推理服务 A\nCPU / GPU · Service', { size: 19, bold: true });
  const serviceB = box(slide, 'service-b', 1123, 832, 196, 83, C.white, '#E3AE6D', 12, false); setText(serviceB, '推理服务 B\nCPU / GPU · Service', { size: 19, bold: true });

  const observePanel = box(slide, 'observability', 1423, 700, 396, 284, C.white, C.orangeLine, 20, true);
  textBox(slide, 'observability-title', '可观测与告警中心', 1455, 714, 330, 38, { size: 27, bold: true, align: 'left' });
  const prometheus = box(slide, 'prometheus', 1457, 765, 328, 66, '#FFF1E3', '#FFF1E3', 13, false); setText(prometheus, 'Prometheus\n集群 / 服务 / 推理指标采集', { size: 19, bold: true });
  const grafana = box(slide, 'grafana', 1457, 846, 158, 91, '#FFF1E3', '#FFF1E3', 13, false); setText(grafana, 'Grafana\n多租户看板', { size: 20, bold: true });
  const alerts = box(slide, 'alert-rules', 1627, 846, 158, 91, '#FFF1E3', '#FFF1E3', 13, false); setText(alerts, 'Alert Rules\n资源 / 延迟\n流量 / 异常', { size: 19, bold: true });

  // Connectors (editable PowerPoint connectors; automatically placed behind nodes)
  connect(slide, user, gateway, { color: C.blue, width: 4 });
  connect(slide, user, auth, { kind: 'elbow', from: 'bottom', to: 'top' });
  connect(slide, source, pipeline);
  connect(slide, pipeline, harbor);
  connect(slide, harbor, configRepo);
  connect(slide, configRepo, argocd);
  connect(slide, argocd, cluster, { kind: 'elbow', from: 'bottom', to: 'top' });
  connect(slide, cluster, observePanel, { color: C.orangeLine, dashed: true });
  connect(slide, alerts, prometheus, { from: 'top', to: 'bottom', color: C.orangeLine, dashed: true });

  // Flow labels
  textBox(slide, 'flow-1', '① 登录 / 创建服务', 470, 247, 240, 28, { size: 17, color: C.text });
  textBox(slide, 'flow-2', '② 构建', 630, 512, 120, 28, { size: 17, color: C.text });
  textBox(slide, 'flow-3', '③ 推送镜像', 958, 512, 150, 28, { size: 17, color: C.text });
  textBox(slide, 'flow-4', '④ 写入声明', 1260, 512, 150, 28, { size: 17, color: C.text });
  textBox(slide, 'flow-5', '⑤ 持续同步', 1602, 512, 150, 28, { size: 17, color: C.text });
  textBox(slide, 'flow-7', '⑦ Metrics', 1363, 872, 115, 28, { size: 17, color: C.text });

  // Security footer
  const security = box(slide, 'security-footer', 205, 1006, 1614, 48, C.security, C.security, 12, false);
  setText(security, '双层安全体系：平台层 JWT + Redis 认证　｜　集群层 Kubernetes RBAC 授权　｜　Namespace 级资源与权限隔离', { size: 21, bold: true, color: C.white });

  slide.speakerNotes.textFrame.setText('[Sources]\n- User-provided project description.\n- Local code: app/service/harbor_pullandpush_service.go (prepareBuildContextTar, dockerBuild).\n- Local code: app/middleware/auth_middleware.go.\n- Local code: app/service/k8s_rbac_service.go.');
  slide.speakerNotes.setVisible(true);

  await writeBlob(PREVIEW, await deck.export({ slide, format: 'png', scale: 1 }));
  await fs.writeFile(LAYOUT, await (await slide.export({ format: 'layout' })).text());
  const pptx = await PresentationFile.exportPptx(deck);
  await pptx.save(OUT);
  console.log(`Wrote ${OUT}`);
}

main().catch((err) => { console.error(err); process.exitCode = 1; });
