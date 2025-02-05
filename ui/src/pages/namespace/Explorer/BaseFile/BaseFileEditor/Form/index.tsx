// import { BaseFileFormSchema, BaseFileFormSchemaType } from "../schema";
// import { Controller, UseFormReturn, useForm, useWatch } from "react-hook-form";
// import React, { FC } from "react";

// import { Fieldset } from "~/components/Form/Fieldset";
// import { FormProps } from "react-router-dom";
// import { zodResolver } from "@hookform/resolvers/zod";

// type BaseFileFormProps = {
//   defaultConfig: BaseFileFormSchemaType;
//   children: (args: {
//     formControls: UseFormReturn<BaseFileFormSchemaType>;
//     formMarkup: JSX.Element;
//     values: BaseFileFormSchemaType;
//   }) => React.ReactNode;
// };

// export const BaseFileForm: FC<FormProps> = ({ defaultConfig, children }) => {
//   const formControls = useForm<BaseFileFormSchemaType>({
//     resolver: zodResolver(BaseFileFormSchema),
//     defaultValues: {
//       ...defaultConfig,
//     },
//   });

//   const fieldsInOrder = BaseFileFormSchema.keyof().options;

//   const watchedValues = useWatch({
//     control: formControls.control,
//   });

//   const values = fieldsInOrder.reduce(
//     (object, key) => ({ ...object, [key]: watchedValues[key] }),
//     {}
//   );

//   const { register, control } = formControls;

//   return children({
//     formControls,
//     values,
//     formMarkup: (
//       <div className="flex flex-col gap-8">
//         <div className="flex gap-3">
//           <Fieldset
//             label={t("pages.explorer.consumer.editor.form.username")}
//             htmlFor="username"
//             className="grow"
//           >
//             <Input {...register("title")} id="title" />
//           </Fieldset>
//           <Fieldset label="Title" htmlFor="title" className="grow">
//             <Input
//               {...register("title", {
//                 setValueAs: treatEmptyStringAsUndefined,
//               })}
//               id="title"
//               type="text"
//             />
//           </Fieldset>
//         </div>
//         <Fieldset label="Version" htmlFor="version">
//           <Input
//             {...register("version", {
//               setValueAs: treatEmptyStringAsUndefined,
//             })}
//             id="version"
//             type="text"
//           />
//         </Fieldset>
//       </div>
//     ),
//   });
// };
