const yamlExtensions = [".yaml", ".yml"] as const;

export const addYamlFileExtension = (name: string) => {
  const newName = name.trim();
  if (yamlExtensions.some((extension) => newName.endsWith(extension))) {
    return newName;
  }
  return `${newName}.yaml`;
};
