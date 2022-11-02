maybe u need learn this: https://learn.microsoft.com/en-us/training/modules/create-nodejs-project-dependencies/


```shell
mkdir <new lambda js handler func name>
npx npm init
# install aws sdk for js on node
npx npm install aws-sdk 
```


OR use sam cli, [local debug lambda](https://docs.aws.amazon.com/zh_cn/serverless-application-model/latest/developerguide/serverless-sam-cli-using-invoke.html)


tips: don't upload dir node_modules, u can do like this: [How to use npm modules in AWS Lambda
](https://bobbyhadz.com/blog/aws-lambda-use-npm-modules)



more help to see sdk for js: https://docs.amazonaws.cn/sdk-for-javascript/index.html