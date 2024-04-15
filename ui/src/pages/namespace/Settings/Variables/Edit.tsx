import {
  VarFormCreateEditSchemaType,
  VarSchemaType,
} from "~/api/variables/schema";

import { DialogContent } from "~/design/Dialog";
import { EditVariableForm } from "~/components/VariableForm/EditForm";
import { useUpdateVar } from "~/api/variables/mutate/update";
import { useVarDetails } from "~/api/variables/query/details";

type EditProps = {
  item: VarSchemaType;
  onSuccess: () => void;
  unallowedNames: string[];
};

const Edit = ({ item, onSuccess, unallowedNames }: EditProps) => {
  const { data, isSuccess } = useVarDetails(item.id);
  const { mutate: updateVar } = useUpdateVar({
    onSuccess,
  });

  const onMutate = (data: VarFormCreateEditSchemaType) => {
    updateVar({
      id: item.id,
      ...data,
    });
  };

  return (
    <DialogContent>
      {isSuccess && (
        <EditVariableForm
          onMutate={onMutate}
          unallowedNames={unallowedNames}
          variable={data.data}
        />
      )}
    </DialogContent>
  );
};

export default Edit;
