import { EnvVarSchemaType } from "~/api/services/schema/services";
import type { Page } from "@playwright/test";
import { PatchSchemaType } from "~/pages/namespace/Explorer/Service/ServiceEditor/schema";
import { createFile } from "e2e/utils/files";

type Service = {
  name: string;
  image: string;
  scale: number;
  size: string;
  cmd: string;
  patches?: PatchSchemaType[];
  envs?: EnvVarSchemaType[];
};

export const createServiceJson = ({
  image,
  scale,
  size,
  cmd,
  patches,
  envs,
}: Service) => {
  // despite the name, this now matches the JSON used in the editor
  const json = {
    image,
    scale,
    size,
    cmd,
    ...(patches && patches.length ? { patches } : {}),
    ...(envs && envs.length ? { envs } : {}),
  };

  return JSON.stringify(json, null, 2);
};

export const createService = async (namespace: string, service: Service) => {
  const content = createServiceJson(service);

  await createFile({
    name: service.name,
    namespace,
    type: "service",
    mimeType: "application/json",
    content,
  });
};

// Helper to access the monaco editor content.
//
// We use the ARIA textbox role instead of the internal
// '.lines-content' class from Monaco. The '.lines-content' element is an
// implementation detail and only reflects the currently visible portion
// of the document, which makes assertions weak (viewport dependent)
// and tightly coupled to Monaco internals. Targeting the labeled textbox
// keeps tests stable across layout changes and updates.

const getServiceEditor = (page: Page) =>
  page.getByRole("textbox", { name: /Editor content/i });

export const getServiceEditorContent = async (page: Page) => {
  const editor = getServiceEditor(page);
  return editor.inputValue();
};
