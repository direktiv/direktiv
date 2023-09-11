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

export const removeYamlFileExtension = (name: string) => {
  const newName = name.trim();

  const endsWithYamlExtension = yamlExtensions.find((extension) =>
    newName.endsWith(extension)
  );

  if (endsWithYamlExtension) {
    return newName.slice(0, -endsWithYamlExtension.length);
  }

  return newName;
};
