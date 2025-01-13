import { existsSync, mkdirSync, readFileSync, writeFileSync } from "fs";

type Item = {
  title: string;
  value: string;
};
export const saveItems = async (
  items: Item[],
  force?: boolean
) => {
  const settingsPath = "./task/result";
  let currentData: { [key: string]: string } = {};

  // Check if settings file exists
  if (existsSync(`${settingsPath}/contract.json`)) {
    const rawData = readFileSync(`${settingsPath}/contract.json`, "utf-8");
    currentData = JSON.parse(rawData);
  }
  // Merge/overwrite the new data
  for (const item of items) {
    currentData[item.title] = item.value;
  }

  // Check if settings directory exists
  if (!existsSync(settingsPath)) {
    mkdirSync(settingsPath, { recursive: true });
  }

  // Save the merged/updated data back to file
  const json = JSON.stringify(currentData, null, 2); // The third parameter makes the output more readable
  writeFileSync(`${settingsPath}/contract.json`, json, "utf-8");
};
