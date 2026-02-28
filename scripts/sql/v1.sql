create table if not exists xtag
(
    id          bigint auto_increment comment 'id'
        primary key,
    label varchar(255) not null comment 'label',
    cipher_text varchar(255) not null comment 'ciphertext',
    created_at  timestamp   default CURRENT_TIMESTAMP not null comment 'creation time',
    updated_at  timestamp   default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment 'update time',
)   comment 'xtag' collate = utf8mb4_bin;

create table if not exists tset
(
    id          bigint auto_increment comment 'id'
        primary key,
    address varchar(255) not null comment 'address',
    value varchar(255) not null comment 'value',
    alpha varchar(255) not null comment 'alpha',
    created_at  timestamp   default CURRENT_TIMESTAMP not null comment 'creation time',
    updated_at  timestamp   default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment 'update time',
    
)   comment 'tset' collate = utf8mb4_bin;