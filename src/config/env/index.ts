import { envVariablesSchema } from "./schema";

export default envVariablesSchema.parse(import.meta.env);
