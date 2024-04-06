create database if not exists webook;
create table if not exists webook.interactives
(
    id           bigint auto_increment
        primary key,
    biz_id       bigint       null,
    biz          varchar(128) null,
    read_cnt     bigint       null,
    favorite_cnt bigint       null,
    like_cnt     bigint       null,
    ctime        bigint       null,
    utime        bigint       null,
    constraint biz_type_id
        unique (biz_id, biz)
);

create table if not exists webook.favorite_items
(
    id     bigint auto_increment
        primary key,
    fid    bigint       null,
    biz_id bigint       null,
    biz    varchar(128) null,
    uid    bigint       null,
    ctime  bigint       null,
    utime  bigint       null,
    constraint biz_type_id_uid
        unique (biz_id, biz, uid)
);

create index idx_favorite_items_fid
    on favorite_items (fid);

create table webook.likes
(
    id     bigint auto_increment
        primary key,
    biz_id bigint           null,
    biz    varchar(128)     null,
    uid    bigint           null,
    status tinyint unsigned null,
    ctime  bigint           null,
    utime  bigint           null,
    constraint biz_type_id_uid
        unique (biz_id, biz, uid)
);
INSERT INTO `interactives`(`biz_id`, `biz`, `read_cnt`, `favorite_cnt`, `like_cnt`, `ctime`, `utime`)
VALUES(1,"test",5725,163,3661,1712165986366,1712165986366),
(2,"test",7216,2688,7006,1712165986366,1712165986366),
(3,"test",8305,9806,1612,1712165986366,1712165986366),
(4,"test",6353,9813,3936,1712165986366,1712165986366),
(5,"test",7289,5993,274,1712165986366,1712165986366),
(6,"test",8661,7973,1403,1712165986366,1712165986366),
(7,"test",4707,2185,3937,1712165986366,1712165986366),
(8,"test",1944,846,7039,1712165986366,1712165986366),
(9,"test",9876,5607,8878,1712165986366,1712165986366),
(10,"test",4029,4180,9166,1712165986366,1712165986366)