const yamlExtensions = [
  ".yaml", // this first one is the default
  ".yml",
] as const;

export const addYamlFileExtension = (name: string) => {
  const newName = name.trim();
  if (yamlExtensions.some((extension) => newName.endsWith(extension))) {
    return newName;
  }
  return `${newName}${yamlExtensions[0]}`;
};

const workflowExtension = ".wf.ts";

export const addWorkflowFileExtension = (name: string) => {
  const newName = name.trim();
  if (newName.endsWith(workflowExtension)) {
    return newName;
  }
  return `${newName}${workflowExtension}`;
};

const serviceExtension = ".svc.json";

export const addServiceFileExtension = (name: string) => {
  const newName = name.trim();
  if (newName.endsWith(serviceExtension)) {
    return newName;
  }
  return `${newName}${serviceExtension}`;
};
