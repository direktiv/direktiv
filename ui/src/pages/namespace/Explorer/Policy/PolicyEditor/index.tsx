import { FC } from "react";
import { FileSchemaType } from "~/api/files/schema";

type PolicyEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const PolicyEditor: FC<PolicyEditorProps> = ({ data }) => <>{data}</>;

export default PolicyEditor;
