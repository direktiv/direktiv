import { FC } from "react";
import { FileSchemaType } from "~/api/files/schema";
import PageEditorPoc from "../poc/PageEditor";

type PageEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const PageEditor: FC<PageEditorProps> = () => <PageEditorPoc />;

export default PageEditor;
