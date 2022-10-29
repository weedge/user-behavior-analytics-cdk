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
    "createAt"      varchar(32),
    "INGREST_ROW_TIME"      varchar(32),
    "APPROXIMATE_ARRIVAL_TIME"      varchar(32)
);

-- Filter errorMsg like warning pump
-- Aggregation with time window(u can use stagger windows,tumbling windows, sliding windows)
-- use tumbling windows for this case
CREATE OR REPLACE PUMP "STREAM_PUMP" AS
    INSERT INTO "DESTINATION_SQL_STREAM"
    SELECT STREAM "eventId", "userId", "objectId", "bizId", "createAt"
        "errorMsg","action",
        STEP("SOURCE_SQL_STREAM_001".ROWTIME BY INTERVAL '60' SECOND) AS "INGREST_ROW_TIME",
        STEP("SOURCE_SQL_STREAM_001".APPROXIMATE_ARRIVAL_TIME BY INTERVAL '60' SECOND) AS "APPROXIMATE_ARRIVAL_TIME",
        -- STEP("SOURCE_SQL_STREAM_001".EVENT_TIME BY INTERVAL '60' SECOND) AS "EVENT_TIME",
        COUNT(*) AS "action_warn_count"
    FROM "SOURCE_SQL_STREAM_001"
    WHERE "errorMsg" LIKE "% WARNNING %" or "errorMsg" LIKE "% warnning %"
    GROUP BY "action",
        STEP("SOURCE_SQL_STREAM_001".ROWTIME BY INTERVAL '60' SECOND),
        STEP("SOURCE_SQL_STREAM_001".APPROXIMATE_ARRIVAL_TIME BY INTERVAL '60' SECOND)
        -- STEP("SOURCE_SQL_STREAM_001".EVENT_TIME BY INTERVAL '60' SECOND) 
    Having "action_warn_count" >= 10;
    