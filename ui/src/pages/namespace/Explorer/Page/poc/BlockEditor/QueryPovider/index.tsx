import {
  QueryProvider as QueryProviderSchema,
  QueryProviderType,
} from "../../schema/blocks/queryProvider";

import { BlockEditFormProps } from "..";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";

type HeadlineEditFormProps = BlockEditFormProps<QueryProviderType>;

export const QueryProvider = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: HeadlineEditFormProps) => {
  const form = useForm<QueryProviderType>({
    resolver: zodResolver(QueryProviderSchema),
    defaultValues: { ...propBlock },
  });

  return (
    <>
      <DialogHeader action={action} path={path} type="headline" />
      <DialogFooter onSubmit={() => onSubmit(form.getValues())} />
    </>
  );
};
