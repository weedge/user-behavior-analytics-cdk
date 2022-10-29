/* CREATE TABLE IN ODS */
/*
######################################
#        FOR JSON SOURCE FILE        #
######################################
*/
CREATE TABLE IF NOT EXISTS ods_raw_event(
  eventId       varchar(64) not null distkey,
  action        varchar(256) not null,
  userId        varchar(64) not null,
  objectId      varchar(64) not null,
  bizId         varchar(64) not null,
  errorMsg      varchar(1024) not null,
  createdAt      varchar(32) not null  sortkey,
  primary key(eventId)
);

/* ADD additional column later if needed */
ALTER TABLE ods_raw_event
ADD column ext varchar(100);

/* Load json format data */
COPY ods_raw_event
FROM 's3://YOUR-S3-BUCKET/***.json' 
IAM_ROLE 'YOUR-REDSHIFT-CLUSTER-IAM-ROLE-ARN'
json 'auto ignorecase';

/* Load json format GZIP compressed data */
COPY ods_raw_event
FROM 's3://YOUR-BUCKET/raw/***.gz' 
iam_role 'YOUR-REDSHIFT-CLUSTER-IAM-ROLE-ARN'
json 'auto ignorecase' 
GZIP ACCEPTINVCHARS TRUNCATECOLUMNS TRIMBLANKS;

/* Load json format PARQUET SNAPPY compressed data */
COPY product_reviews_parquet
FROM 's3://YOUR-BUCKET/****.parquet.snappy' 
IAM_ROLE 'YOUR-REDSHIFT-CLUSTER-IAM-ROLE-ARN'
FORMAT AS PARQUET;


/* Start to analytics */
SELECT count(*),action FROM "dev"."public"."ods_raw_event" 
WHERE errormsg = '[error]'
GROUP BY action;
