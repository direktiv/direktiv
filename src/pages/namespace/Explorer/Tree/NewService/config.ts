import { RJSFSchema } from "@rjsf/utils";
import { stringify } from "json-to-pretty-yaml";
import { useTranslation } from "react-i18next";

export const useServiceFormSchema = (): RJSFSchema => {
  const { t } = useTranslation();
  return {
    properties: {
      image: {
        title: t("pages.explorer.tree.newService.form.image"),
        type: "string",
      },
      scale: {
        title: t("pages.explorer.tree.newService.form.scale"),
        type: "integer",
        enum: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
      },
      size: {
        title: t("pages.explorer.tree.newService.form.size"),
        type: "integer",
        enum: ["large", "medium", "small"],
      },
      cmd: {
        title: t("pages.explorer.tree.newService.form.cmd"),
        type: "string",
      },
    },
    required: ["image", "name"],
    type: "object",
  };
};

export const serviceHeader = {
  direktiv_api: "service/v1",
};

export const addServiceHeader = (serviceJSON: object) => ({
  ...serviceHeader,
  ...serviceJSON,
});

export const defaultServiceYaml = stringify(serviceHeader);
