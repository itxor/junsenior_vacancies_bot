create table vacancies
(
    id              varchar(30)               not null,
    name            varchar(255) charset utf8 not null,
    place           varchar(255) charset utf8 not null,
    salary_from     int                       null,
    salary_to       int                       null,
    salary_currency varchar(255) charset utf8 null,
    salary_gross    bit default b'0'          null,
    publiched_at    varchar(40) charset utf8  not null,
    archived        bit default b'0'          null,
    url             varchar(500) charset utf8 not null,
    employer_name   varchar(255) charset utf8 null,
    constraint id
        unique (id)
);

alter table vacancies
    add primary key (id);