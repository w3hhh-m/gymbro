drop table if exists exercisemusclegroups cascade;

drop table if exists musclegroups cascade;

alter table clans
    drop constraint if exists fk_owner_id cascade;

drop table if exists userexercisemaxweights cascade;

drop table if exists records cascade;

drop table if exists exercises cascade;

drop table if exists workouts cascade;

drop table if exists subscriptions cascade;

drop table if exists users cascade;

drop table if exists gyms cascade;

drop table if exists clans cascade;

