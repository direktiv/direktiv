import Form from "@rjsf/core";

export default {
  title: "Components/JSON schema Form",
};

// for examples to to
// https://rjsf-team.github.io/react-jsonschema-form/
// select one template copy the generated JSONSchemach

export const Default = () => (
  <Form
    schema={{
      title: "A registration form",
      type: "object",
      required: ["firstName", "lastName"],
      properties: {
        password: {
          type: "string",
          title: "Password",
        },
        lastName: {
          type: "string",
          title: "Last name",
        },
        bio: {
          type: "string",
          title: "Bio",
        },
        firstName: {
          type: "string",
          title: "First name",
        },
        age: {
          type: "integer",
          title: "Age",
        },
      },
    }}
  />
);
