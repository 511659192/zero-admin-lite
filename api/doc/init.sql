CREATE TABLE `sys_user` (
                            `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
                            `mobile` char(11) NOT NULL DEFAULT '' COMMENT '手机',
                            `user_name` varchar(50) NOT NULL COMMENT '姓名',
                            `password` varchar(64) CHARACTER SET utf8mb3 COLLATE utf8mb3_general_ci NOT NULL COMMENT '密码',
                            `status_id` tinyint NOT NULL DEFAULT '1' COMMENT '状态(1:正常，0:禁用)',
                            `sort` int NOT NULL DEFAULT '1' COMMENT '排序',
                            `remark` varchar(255) DEFAULT NULL COMMENT '备注',
                            `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `AK_phone` (`mobile`)
    ) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_0900_ai_ci COMMENT='用户信息';


-- CREATE TABLE `sys_role` (
--                             `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
--                             `role_name` varchar(50) NOT NULL COMMENT '名称',
--                             `status_id` tinyint NOT NULL DEFAULT '1' COMMENT '状态(1:正常，0:禁用)',
--                             `sort` int NOT NULL DEFAULT '1' COMMENT '排序',
--                             `remark` varchar(255) NOT NULL COMMENT '备注',
--                             `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
--                             `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
--     PRIMARY KEY (`id`),
--     UNIQUE KEY `role_name` (`role_name`),
--     KEY `name_status_index` (`role_name`,`status_id`)
--     ) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE =utf8mb4_0900_ai_ci COMMENT='角色信息';
