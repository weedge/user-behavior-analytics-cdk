"""Producer produces fake data to be inputted into a Kinesis stream."""

import json
import time
import uuid
import random
from datetime import datetime
from pprint import pprint

import boto3

from faker import Faker

# This boots up the kinesis analytic application so you don't have to click "run" on the kinesis analytics console
try:
    kinesisanalytics = boto3.client(
        "kinesisanalyticsv2", region_name="us-east-1")
    # aws kinesisanalyticsv2 start-application --application-name abnormality-event-detector --run-configuration { RUN-CONF }
    kinesisanalytics.start_application(
        ApplicationName="abnormality-event-detector",
        RunConfiguration={
            'SqlRunConfigurations': [
                {
                    'InputId': '1.1',
                    'InputStartingPositionConfiguration': {
                        'InputStartingPosition': 'NOW'
                    }
                },
            ]
        }
    )
    print("Giving 30 seconds for the kinesis analytics application to boot")
    time.sleep(30)
except kinesisanalytics.exceptions.ResourceInUseException:
    print("Application already running, skipping start up step")

eventSteamName = input(
    "Please enter the stream name that was outputted from cdk deploy - (StreamingSolutionWithCdkStack.EventStreamName): ")
kinesis = boto3.client("kinesis", region_name="us-east-1")
fake = Faker()

i = 1
total = 10
while True:
    if i > total:
        break
    payload = {
        "eventId": str(uuid.uuid4()),
        "action": fake.name(),
        "userId": str(uuid.uuid4()),
        "objectId": str(uuid.uuid4()),
        "bizId": str(uuid.uuid4()),
        # "errorMsg": fake.name(),
        "errorMsg": "[error]",
        "createdAt": str(datetime.now()),
    }
    pprint(payload)
    response = kinesis.put_record(
        StreamName=eventSteamName, Data=json.dumps(payload), PartitionKey="eventId"
    )
    i = i+1
    pprint(response)
    time.sleep(3)
