import { CreateVariableForm } from "../../../../../../components/VariableForm/CreateForm";
import { VarFormCreateEditSchemaType } from "~/api/variables/schema";
import { useCreateVar } from "~/api/variables/mutate/create";

type CreateProps = {
  onSuccess: () => void;
  path: string;
  unallowedNames: string[];
};

const Create = ({ onSuccess, path, unallowedNames }: CreateProps) => {
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
    <CreateVariableForm onMutate={onMutate} unallowedNames={unallowedNames} />
  );
};

export default Create;
