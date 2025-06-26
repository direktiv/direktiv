import { BlockForm } from "..";
import { usePageEditor } from "../../PageCompiler/context/pageCompilerContext";

export const EditorPanel = () => {
  const { focus } = usePageEditor();

  return <>{focus && <BlockForm action="edit" path={focus} />}</>;
};
