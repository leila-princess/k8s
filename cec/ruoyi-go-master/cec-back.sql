
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for k8s_namespace
-- ----------------------------
DROP TABLE IF EXISTS `k8s_namespace`;
CREATE TABLE `k8s_namespace`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '命名空间名称',
  `user_id` int(11) NOT NULL COMMENT '所属用户ID',
  `status` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `service_account` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT 'ServiceAccount名称',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_name`(`name`) USING BTREE,
  INDEX `idx_user_id`(`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'Kubernetes命名空间表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of k8s_namespace
-- ----------------------------
INSERT INTO `k8s_namespace` VALUES (1, 'test-tenant', 2, '0', 'test-tenant-sa', '2025-08-29 09:30:44', '2025-08-29 09:30:45');
INSERT INTO `k8s_namespace` VALUES (2, 'test1', 3, '0', 'test1-sa', '2025-08-29 13:35:28', '2025-08-29 13:35:28');
INSERT INTO `k8s_namespace` VALUES (3, 'test22', 4, '0', 'test22-sa', '2025-09-01 09:22:45', '2025-09-01 15:40:35');
INSERT INTO `k8s_namespace` VALUES (7, 'test66', 8, '1', 'test66-sa', '2025-09-02 14:07:53', '2025-09-02 14:20:58');
INSERT INTO `k8s_namespace` VALUES (8, 'test777', 9, '0', 'test777-sa', '2025-09-02 14:34:28', '2025-09-02 14:34:29');

-- ----------------------------
-- Table structure for k8s_resource_quota
-- ----------------------------
DROP TABLE IF EXISTS `k8s_resource_quota`;
CREATE TABLE `k8s_resource_quota`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `namespace_id` int(11) NOT NULL COMMENT '命名空间ID',
  `cpu_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'CPU限制',
  `cpu_request` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'CPU请求',
  `memory_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '内存限制',
  `memory_request` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '内存请求',
  `gpu_limit` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'GPU限制',
  `gpu_memory` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'GPU显存限制',
  `status` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_namespace_id`(`namespace_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = 'Kubernetes资源配额表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of k8s_resource_quota
-- ----------------------------
INSERT INTO `k8s_resource_quota` VALUES (1, 1, '3', '1', '3Gi', '2Gi', '1', '8Gi', '0', '2025-08-29 09:30:44', '2025-09-01 17:37:22');
INSERT INTO `k8s_resource_quota` VALUES (2, 2, '1000m', '500m', '2Gi', '1Gi', '1', '8Gi', '0', '2025-08-29 13:35:28', '2025-08-29 13:35:28');
INSERT INTO `k8s_resource_quota` VALUES (3, 3, '1000m', '500m', '2Gi', '1Gi', '1', '8Gi', '0', '2025-09-01 09:22:45', '2025-09-01 15:40:35');
INSERT INTO `k8s_resource_quota` VALUES (7, 7, '3', '1', '3Gi', '2Gi', '1', '8Gi', '1', '2025-09-02 14:07:53', '2025-09-02 14:21:45');
INSERT INTO `k8s_resource_quota` VALUES (8, 8, '1000m', '500m', '2Gi', '1Gi', '1', '8Gi', '0', '2025-09-02 14:34:29', '2025-09-02 14:34:29');

-- ----------------------------
-- Table structure for tenant_user
-- ----------------------------
DROP TABLE IF EXISTS `tenant_user`;
CREATE TABLE `tenant_user`  (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '登录用户名',
  `password` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '登录密码',
  `status` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '状态（0正常 1停用）',
  `is_admin` char(1) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '0' COMMENT '是否管理员（0否 1是）',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `idx_username`(`username`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '租户用户表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tenant_user
-- ----------------------------
INSERT INTO `tenant_user` VALUES (2, 'test-tenant', '$2a$10$F7f797r9oGJ/iW6t4GU1uuN9e8F3nekn7CF7aacMtrikBiiRWvnCq', '0', '0', '2025-08-29 09:30:44', '2025-08-29 09:30:44');
INSERT INTO `tenant_user` VALUES (3, 'test1', '$2a$10$/6b289OLML8upFcFiY1Ms.UmeRibFPh6LlPKGL7OpJGeqWOHO2Ehe', '0', '1', '2025-08-29 13:35:28', '2025-09-01 13:52:08');
INSERT INTO `tenant_user` VALUES (4, 'test22', '$2a$10$hWruhiBsisZw1qpf8pcvUerXhKGFW1NAs2Ih57QG.cSeXXAvZdivG', '0', '0', '2025-09-01 09:22:43', '2025-09-02 11:33:46');
INSERT INTO `tenant_user` VALUES (8, 'test66', '$2a$10$987gfaDAyNDDW8CyYN970.Q3RWusTk7bwRfXlSfHY8sX4AuKb5.X.', '1', '0', '2025-09-02 14:07:53', '2025-09-02 14:20:59');
INSERT INTO `tenant_user` VALUES (9, 'test777', '$2a$10$ZBPx9eOf1NP.In2FCmsZ0e0syiaYBRC97vnECl12qBE6VhRnb26te', '0', '0', '2025-09-02 14:34:28', '2025-09-02 14:34:28');

SET FOREIGN_KEY_CHECKS = 1;

CREATE TABLE `image_task` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tenant_id` int(11) NOT NULL COMMENT '租户ID',
  `task_type` varchar(20) NOT NULL COMMENT '任务类型：push/pull',
  `image_name` varchar(255) NOT NULL COMMENT '镜像名称',
  `tag` varchar(100) NOT NULL COMMENT '标签',
  `status` varchar(20) NOT NULL COMMENT '任务状态',
  `message` text COMMENT '任务消息',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='镜像任务表';

ALTER TABLE image_task
ADD COLUMN file_path VARCHAR(255);



CREATE TABLE `argocd_deployment` (
                                     `id` int NOT NULL AUTO_INCREMENT,
                                     `tenant_id` int NOT NULL,
                                     `name` varchar(191) NOT NULL,
                                     `git_repo` varchar(191) NOT NULL,
                                     `git_branch` varchar(191) NOT NULL,
                                     `yaml_path` varchar(191) NOT NULL,
                                     `image` varchar(191) NOT NULL,
                                     `trigger_type` varchar(191) NOT NULL,
                                     `schedule` varchar(191),
                                     `status` varchar(191) NOT NULL,
                                     `last_run` bigint,
                                     `created_at` bigint,
                                     `updated_at` bigint,
                                     PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
ALTER TABLE `argocd_deployment` RENAME TO `argo_cd_deployment`;

CREATE TABLE ns_harbor_secrets (
                                   namespace VARCHAR(100) PRIMARY KEY,
                                   secret_value TEXT NOT NULL
);