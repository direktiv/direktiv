export const workflowWithSecrets = `direktiv_api: workflow/v1
functions:
- id: aws-cli
  image: direktiv/aws-cli:dev
  type: knative-workflow

states:
- id: start-instance
  type: action
  action:
    secrets: ["ACCESS_KEY", "ACCESS_SECRET"]
    function: aws-cli
    input: 
      access-key: jq(.secrets.ACCESS_KEY)
      secret-key: jq(.secrets.ACCESS_SECRET)
      region: ap-southeast-2
      commands: 
      - command: aws ec2 run-instances --image-id ami-07620139298af599e --instance-type t2.small
`;
