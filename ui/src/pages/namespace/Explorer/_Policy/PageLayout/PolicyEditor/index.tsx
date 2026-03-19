import { FC } from "react";
import { FileSchemaType } from "~/api/files/schema";
import { PlaceholderPolicy } from "./PlaceholderPolicy";

type PolicyEditorProps = {
  data: FileSchemaType;
};

export const PolicyEditor: FC<PolicyEditorProps> = ({ data }) => (
  // To Do: if data is empty, show placeholder
  // but for that we need the correct type for the cedar policy file
  // For now, we will always show the placeholder

  <>
    Policy content is: {JSON.stringify(data.data)} <PlaceholderPolicy />
  </>
);
