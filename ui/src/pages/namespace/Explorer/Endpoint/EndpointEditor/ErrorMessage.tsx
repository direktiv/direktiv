import { EndpointFormSchemaType } from "./schema";
import { FieldErrors } from "react-hook-form";
import FormErrors from "~/components/FormErrors";

type ErrorMessageProps = {
  errors: FieldErrors<EndpointFormSchemaType>;
};
export const ErrorMessage = ({ errors }: ErrorMessageProps) => {
  let errorsToDisplay = errors;
  /**
   * to avoid having unspecified error messages, we check for the common
   * error key that the user can have in the form, and if it exists, we
   * use it instead of the generic error key from the root
   */
  if (errors["x-direktiv-config"]) {
    errorsToDisplay = errors["x-direktiv-config"];
    if (errors["x-direktiv-config"]["plugins"]) {
      errorsToDisplay = errors["x-direktiv-config"]["plugins"];
    }
  }

  return <FormErrors errors={errorsToDisplay} className="mb-5" />;
};
