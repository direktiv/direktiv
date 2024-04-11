import { CreateVariableForm } from "../../../../../../components/VariableForm/CreateForm";
import { VarFormCreateEditSchemaType } from "~/api/variables/schema";
import { useCreateVar } from "~/api/variables/mutate/create";
import { useTranslation } from "react-i18next";

type CreateProps = {
  onSuccess: () => void;
  path: string;
  unallowedNames: string[];
};

const Create = ({ onSuccess, path, unallowedNames }: CreateProps) => {
  const { t } = useTranslation();
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  const onMutate = (data: VarFormCreateEditSchemaType) => {
    createVar({
      ...data,
      workflowPath: path,
    });
  };

  return (
    <CreateVariableForm
      title={t("pages.explorer.tree.workflow.settings.variables.create.title")}
      onMutate={onMutate}
      unallowedNames={unallowedNames}
    />
  );
};

export default Create;
