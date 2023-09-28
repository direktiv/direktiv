import { z } from "zod";

/**
 * example:
  [
    "opaManage",
    "variablesView",
    "registriesManage",
    "explorerManage",
    "registriesView",
    "nsconfigView",
    "eventsSend",
    "instancesView",
    "secretsView",
    "secretsManage",
    "servicesView",
    "servicesManage",
    "instancesManage",
    "explorerView",
    "workflowView",
    "workflowManage",
    "variablesManage",
    "nsconfigManage",
    "deleteNamespace",
    "eventsView",
    "workflowExecute",
    "workflowStore",
    "permissionsView",
    "permissionsManage",
    "opaView",
    "eventsManage"
  ]
 */
export const PermissionKeysSchema = z.array(z.string());
