package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsredshift"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RedshiftQuicksightCdkStackProps struct {
	awscdk.StackProps
}

func NewRedshiftQuicksightCdkStack(scope constructs.Construct, id string, props *RedshiftQuicksightCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// create vpc for deploy redshift cluster
	vpc := awsec2.NewVpc(stack, jsii.String("RedshiftVpc"), &awsec2.VpcProps{
		//Cidr:               jsii.String("10.10.0.0/16"),
		IpAddresses:        awsec2.IpAddresses_Cidr(jsii.String("10.10.0.0/16")),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		MaxAzs:             jsii.Number(2),
		NatGateways:        jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{Name: jsii.String("public"), SubnetType: awsec2.SubnetType_PUBLIC, CidrMask: jsii.Number(24)},
			{Name: jsii.String("db"), SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED, CidrMask: jsii.Number(24)},
		},
	})

	// create redshift cluster
	// IAM role
	rsClusterRole := awsiam.NewRole(stack, jsii.String("RedshiftClusterRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("redshift.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonS3ReadOnlyAccess")),
		},
	})

	// subnet group for redshift cluster
	subnetGroup := awsredshift.NewCfnClusterSubnetGroup(stack, jsii.String("RedshiftSubnetGroup"), &awsredshift.CfnClusterSubnetGroupProps{
		Description: jsii.String("redshift subnet group"),
		SubnetIds: vpc.SelectSubnets(&awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		}).SubnetIds,
	})

	// create security group for QuickSight
	quickSightToRedshiftSg := awsec2.NewSecurityGroup(stack, jsii.String("RedshiftSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		Description:       jsii.String("Security Group for QuickSight"),
		SecurityGroupName: jsii.String("RedshiftQuickSightSecurityGroup"),
	})

	// add ingress rule for quicksight
	// https://docs.aws.amazon.com/quicksight/latest/user/regions.html
	quickSightToRedshiftSg.AddIngressRule(
		awsec2.Peer_Ipv4(jsii.String("52.23.63.224/27")),
		awsec2.Port_Tcp(jsii.Number(5439)),
		jsii.String("Allow QuickSight connections"), nil)

	// create cluster password
	secret := awssecretsmanager.NewSecret(stack, jsii.String("SetRedShiftClusterSecret"), &awssecretsmanager.SecretProps{
		Description:   jsii.String("Redshift cluster secret"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		SecretName:    jsii.String("RedshiftClusterSecret"),
		GenerateSecretString: &awssecretsmanager.SecretStringGenerator{
			ExcludePunctuation:   jsii.Bool(true),
			GenerateStringKey:    jsii.String("password"),
			IncludeSpace:         jsii.Bool(false),
			PasswordLength:       jsii.Number(16),
			SecretStringTemplate: stack.ToJsonString(map[string]interface{}{"username": "Administrator"}, nil),
		},
	})

	defPassword := awscdk.NewCfnParameter(stack, jsii.String("password"), &awscdk.CfnParameterProps{
		Default:     jsii.String("Admin1234"),
		Description: jsii.String("Redshift superuser password"),
		NoEcho:      jsii.Bool(true),
		Type:        jsii.String("String"),
	})

	// create redshift cluster, just for dev test, use single-node
	redshiftCluster := awsredshift.NewCfnCluster(stack, jsii.String("RedshiftDemo"), &awsredshift.CfnClusterProps{
		ClusterType:        jsii.String("single-node"),
		DbName:             jsii.String("user_behavior"),
		MasterUsername:     jsii.String("dwh_master"),
		MasterUserPassword: defPassword.ValueAsString(),
		//MasterUserPassword:     secret.SecretValue().ToString(),
		NodeType:               jsii.String("dc2.large"), //"dc2.large","dc2.8xlarge","ra3.xlplus","ra3.4xlarge","ra3.16xlarge"
		ClusterSubnetGroupName: subnetGroup.Ref(),
		IamRoles:               &[]*string{rsClusterRole.RoleArn()},
		NumberOfNodes:          jsii.Number(1), // only 1 for dev test
		Port:                   jsii.Number(5439),
		VpcSecurityGroupIds:    &[]*string{quickSightToRedshiftSg.SecurityGroupId()},
	})

	// output
	awscdk.NewCfnOutput(stack, jsii.String("StackRepoFrom"), &awscdk.CfnOutputProps{
		Value:       jsii.String("https://github.com/weedge/user-behavior-analytics-cdk"),
		Description: jsii.String("how to use this stack, see readme or github page"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("RedshiftCluster"), &awscdk.CfnOutputProps{
		Value:       redshiftCluster.AttrEndpointAddress(),
		Description: jsii.String("Redshift Endpoint"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("RedshiftPasswordKMS"), &awscdk.CfnOutputProps{
		Value: jsii.String("https://console.aws.amazon.com/secretsmanager/home?region=" +
			*awscdk.Aws_REGION() +
			"#/secret?name=" +
			*secret.SecretArn()),
		Description: jsii.String("Redshift KMS password"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("RedshiftIAMRole"), &awscdk.CfnOutputProps{
		Value:       rsClusterRole.RoleArn(),
		Description: jsii.String("Redshift IAM Role Arn"),
	})

	return stack
}
