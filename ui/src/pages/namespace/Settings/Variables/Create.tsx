import { CreateVariableForm } from "~/components/VariableForm/CreateForm";
import { useCreateVar } from "~/api/variables/mutate/create";
import { useTranslation } from "react-i18next";

type CreateProps = { onSuccess: () => void; unallowedNames: string[] };

const Create = ({ onSuccess, unallowedNames }: CreateProps) => {
  const { t } = useTranslation();
  const { mutate: createVar } = useCreateVar({
    onSuccess,
  });

  return (
    <CreateVariableForm
      title={t("pages.settings.variables.create.title")}
      onMutate={createVar}
      unallowedNames={unallowedNames}
    />
  );
};

export default Create;
