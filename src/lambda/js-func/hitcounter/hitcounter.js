const { DynamoDB, Lambda, Kinesis } = require('aws-sdk');
//import { v4 as uuidv4 } from 'uuid';
//import { faker } from '@faker-js/faker';

exports.handler = async function(event) {
  console.log("request:", JSON.stringify(event, undefined, 2));

  // create AWS SDK clients
  const dynamo = new DynamoDB();
  const lambda = new Lambda();

  // update dynamo entry for "path" with hits++
  await dynamo.updateItem({
    TableName: process.env.HITS_TABLE_NAME,
    Key: { path: { S: event.path } },
    UpdateExpression: 'ADD hits :incr',
    ExpressionAttributeValues: { ':incr': { N: '1' } }
  }).promise();

  // notice: put event item to kinesis data streams  or async call downstream function
  // method is POST and path is /event
  if (event.httpMethod == "POST"
    && event.path == "/event"
    && process.env.HITS_STREAM_NAME.length>0) {
    putRecordToKDS(event.body);
  }

  // call downstream function and capture response
  const resp = await lambda.invoke({
    FunctionName: process.env.DOWNSTREAM_FUNCTION_NAME,
    Payload: JSON.stringify(event)
  }).promise();

  console.log('downstream response:', JSON.stringify(resp, undefined, 2));

  // return response back to upstream caller
  return JSON.parse(resp.Payload);
};

async function putRecordToKDS(bodyStr){
  const kinesis = new Kinesis();
  var bodyObj = JSON.parse(bodyStr)
  var recordData = {
    eventId: bodyObj.eventId?bodyObj.eventId:"",
    action: bodyObj.action?bodyObj.action:"",
    userId: bodyObj.userId?bodyObj.userId:"",
    objectId: bodyObj.objectId?bodyObj.objectId:"",
    bizId: bodyObj.bizId?bodyObj.bizId:"",
    errorMsg: bodyObj.errorMsg?bodyObj.errorMsg:"",
    createdTime: new Date()
  };
  console.log('put KDS recordData:', JSON.stringify(recordData, undefined, 2));
  const putRes = await kinesis.putRecord({
    Data: JSON.stringify(recordData),
    StreamName:process.env.HITS_STREAM_NAME,
    PartitionKey:recordData.eventId
  }).promise();
  console.log('KDS result:', JSON.stringify(putRes, undefined, 2));
}

