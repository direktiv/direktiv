export const validationAsFirstState = `description: A simple 'validate' state workflow that checks an email
states:
- id: validate-email
  type: validate
  subject: jq(.)
  schema:
    type: object
    properties:
      email:
        type: string
        format: email
  catch:
  - error: direktiv.schema.*
    transition: email-not-valid 
  transition: email-valid
- id: data
  type: noop
  transform:
    email: "trent.hilliam@direktiv.io"
  transition: validate-email
- id: email-not-valid
  type: noop
  transform:
    result: "Email is not valid."
- id: email-valid
  type: noop
  transform:
    result: "Email is valid."`;

export const validationAsSecondState = `description: A simple 'validate' state workflow that checks an email
states:
- id: data
  type: noop
  transform:
    email: "trent.hilliam@direktiv.io"
  transition: validate-email
- id: validate-email
  type: validate
  subject: jq(.)
  schema:
    type: object
    properties:
      email:
        type: string
        format: email
  catch:
  - error: direktiv.schema.*
    transition: email-not-valid 
  transition: email-valid
- id: email-not-valid
  type: noop
  transform:
    result: "Email is not valid."
- id: email-valid
  type: noop
  transform:
    result: "Email is valid."`;

export const complexValidationAsFirstState = `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: input
  type: validate
  schema:
    title: A registration form
    type: object
    required:
    - firstName
    - lastName
    properties:
      password:
        type: string
        title: Password
      lastName:
        type: string
        title: Last name
      bio:
        type: string
        title: Bio
      firstName:
        type: string
        title: First name
      age:
        type: integer
        title: Age`;

export const noValidateState = `description: A simple 'validate' state workflow that checks an email
states:
- id: data
  type: noop
  transform:
    email: "trent.hilliam@direktiv.io"
  transition: validate-email
- id: email-not-valid
  type: noop
  transform:
    result: "Email is not valid."
- id: email-valid
  type: noop
  transform:
    result: "Email is valid."`;
