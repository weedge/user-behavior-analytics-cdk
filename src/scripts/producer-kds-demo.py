import datetime
import json
import random
import boto3
import time
from pprint import pprint

STREAM_NAME = "kda-input-stream"


def get_data():
    return {
        'event_time': datetime.datetime.now().isoformat(),
        'ticker': random.choice(['AAPL', 'AMZN', 'MSFT', 'INTC', 'TBV']),
        'price': round(random.random() * 100, 2)}


def generate(stream_name, kinesis_client):
    while True:
        data = get_data()
        print(data)
        response = kinesis_client.put_record(
            StreamName=stream_name,
            Data=json.dumps(data),
            PartitionKey="partitionkey")

        pprint(response)
        time.sleep(1)


if __name__ == '__main__':
    session = boto3.Session(profile_name='default')
    generate(STREAM_NAME, session.client('kinesis', region_name='us-east-1'))
