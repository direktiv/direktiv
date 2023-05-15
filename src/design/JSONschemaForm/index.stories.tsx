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
      type: "object",
      required: ["expressions"],
      properties: {
        expressions: {
          type: "array",
          description: "expressions to solve",
          title: "Expressions",
          items: {
            type: "string",
          },
        },
      },
    }}
  />
);
