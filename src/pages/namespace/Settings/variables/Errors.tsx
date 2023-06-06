import Alert from "~/design/Alert";
import { FieldErrors } from "react-hook-form";

const Errors = ({ errors }: { errors: FieldErrors }) => {
  const entries = Object.entries(errors);

  return entries.length ? (
    <Alert variant="error" className="mb-5">
      <ul>
        {entries.map(([field, message]) => (
          <li key={field}>{`${field}: ${message?.message}`}</li>
        ))}
      </ul>
    </Alert>
  ) : (
    <></>
  );
};

export default Errors;
