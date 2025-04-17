import { ApiErrorSchemaType } from "~/api/errorHandling";
import ErrorPage from "./ErrorPage";

const NotFoundPage = () => {
  const error: Error & ApiErrorSchemaType = {
    name: "UI route or resource not found",
    message: "UI route or resource not found",
    status: 404,
  };

  return <ErrorPage error={error}></ErrorPage>;
};

export default NotFoundPage;
