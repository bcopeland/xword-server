create table puzzle (
    id varchar(64) not null primary key,
    type varchar(80) not null default '',
    title varchar(80) not null default '',
    author varchar(80) not null default '',
    editor varchar(80) not null default '',
    copyright varchar(80) not null default '',
    publisher varchar(80) not null default '',
    date varchar(20) not null default '',
    height int not null default 0,
    width int not null default 0,
    grid text not null default '',
    notepad text not null default ''
) engine=InnoDB default charset=utf8;

create table clue (
    id bigint not null auto_increment primary key,
    puzzle_id varchar(64) not null,
    row int not null default 0,
    col int not null default 0,
    number int not null default 0,
    direction tinyint not null default 0,
    answer varchar(255) not null default '',
    text varchar(255) not null default '',
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;

create table rebus (
    id bigint not null auto_increment primary key,
    puzzle_id varchar(64) not null,
    row int not null default 0,
    col int not null default 0,
    short varchar(255) not null default '',
    expanded varchar(255) not null default '',
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;

create table circle (
    id bigint not null auto_increment primary key,
    puzzle_id varchar(64) not null,
    row int not null default 0,
    col int not null default 0,
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;

create table shade (
    id bigint not null auto_increment primary key,
    puzzle_id varchar(64) not null,
    row int not null default 0,
    col int not null default 0,
    color varchar(20) not null default '',
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;

create table shade (
    id bigint not null auto_increment primary key,
    puzzle_id varchar(64) not null,
    row int not null default 0,
    col int not null default 0,
    color varchar(20) not null default '',
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;

create table solution (
    id varchar(64) not null,
    puzzle_id varchar(64) not null,
    version int not null default 0,
    grid text not null default '',
    constraint foreign key (puzzle_id) references puzzle (id) on delete cascade on update cascade
) engine=InnoDB default charset=utf8;
