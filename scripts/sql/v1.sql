create table if not exists xtag
(
    id          bigint auto_increment comment 'id'
        primary key,
    label varchar(255) not null comment '标签',
    cipher_text varchar(255) not null comment '密文',
    created_at  timestamp   default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at  timestamp   default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
)   comment 'xtag' collate = utf8mb4_bin;

create table if not exists tset
(
    id          bigint auto_increment comment 'id'
        primary key,
    address varchar(255) not null comment '地址',
    value varchar(255) not null comment '值',
    alpha varchar(255) not null comment 'alpha',
    created_at  timestamp   default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_at  timestamp   default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    
)   comment 'tset' collate = utf8mb4_bin;