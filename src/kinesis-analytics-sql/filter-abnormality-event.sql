-- ** Continuous Filter ** 
    -- Performs a continuous filter based on a WHERE condition.
    --          .----------.   .----------.   .----------.              
    --          |  SOURCE  |   |  INSERT  |   |  DESTIN. |              
    -- Source-->|  STREAM  |-->| & SELECT |-->|  STREAM  |-->Destination
    --          |          |   |  (PUMP)  |   |          |              
    --          '----------'   '----------'   '----------'               
    -- STREAM (in-application): a continuously updated entity that you can SELECT from and INSERT into like a TABLE
    -- PUMP: an entity used to continuously 'SELECT ... FROM' a source STREAM, and INSERT SQL results into an output STREAM
    -- Create output stream, which can be used to send to a destination
-- reference: 
-- https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/sqlref/analytics-sql-reference.html
-- https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/dev/streaming-sql-concepts.html
-- https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/sqlref/kinesis-analytics-sqlref.pdf

-- abnormality event stream
CREATE OR REPLACE STREAM "DESTINATION_SQL_STREAM" 
(
    "eventId"       varchar(64),
    "action"        varchar(256),
    "userId"        varchar(64),
    "objectId"      varchar(64),
    "bizId"         varchar(64),
    "errorMsg"      varchar(1024),
    "createdAt"      varchar(32)
);

-- Filter errorMsg like panic/error pump
CREATE OR REPLACE PUMP "ERROR_PANIC_STREAM_PUMP" AS
    INSERT INTO "DESTINATION_SQL_STREAM"
    SELECT STREAM "eventId", "action", "userId", "objectId", "bizId", "errorMsg","createdAt"
    FROM "SOURCE_SQL_STREAM_001"
    WHERE "errorMsg" LIKE '%[PANIC]%'
        or "errorMsg" LIKE '%[panic]%' 
        or "errorMsg" LIKE '%[ERROR]%' 
        or "errorMsg" LIKE '%[error]%';