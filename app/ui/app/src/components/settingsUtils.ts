import { Settings as SettingsType } from "@/gotypes";
import type { CloudStatusSource } from "@/api";

export interface SettingsDefaultsActions {
  updateSettings: (settings: SettingsType) => Promise<unknown>;
  updateCloud: (enabled: boolean) => Promise<unknown>;
  updateShowAppsInMenu: (visible: boolean) => Promise<unknown>;
  resetChatGPTModels: () => Promise<boolean>;
  resetClaudeMappings: () => Promise<boolean>;
  currentSettings: SettingsType;
  currentShowAppsInMenu: boolean;
  cloudSource: CloudStatusSource;
  onSaved: () => void;
}

export async function applySettingsDefaults({
  updateSettings,
  updateCloud,
  updateShowAppsInMenu,
  resetChatGPTModels,
  resetClaudeMappings,
  currentSettings,
  currentShowAppsInMenu,
  cloudSource,
  onSaved,
}: SettingsDefaultsActions): Promise<void> {
  const cloudNeedsReset = cloudSource === "config" || cloudSource === "both";
  const rollbacks: Array<() => Promise<unknown>> = [];

  try {
    if (cloudNeedsReset) {
      await updateCloud(true);
      rollbacks.push(() => updateCloud(false));
    }

    await updateSettings(
      new SettingsType({
        Expose: false,
        Browser: false,
        Models: "",
        Agent: false,
        Tools: false,
        ContextLength: currentSettings.ContextLength,
        AutoUpdateEnabled: true,
      }),
    );
    rollbacks.push(() => updateSettings(currentSettings));

    await updateShowAppsInMenu(true);
    rollbacks.push(() => updateShowAppsInMenu(currentShowAppsInMenu));

    if (!(await resetChatGPTModels())) {
      throw new Error("ChatGPT models could not be reset");
    }
    if (!(await resetClaudeMappings())) {
      throw new Error("Claude model mappings could not be reset");
    }
  } catch (error) {
    const rollbackErrors: unknown[] = [];
    for (const rollback of rollbacks.reverse()) {
      try {
        await rollback();
      } catch (rollbackError) {
        rollbackErrors.push(rollbackError);
      }
    }
    if (rollbackErrors.length > 0) {
      console.error("Failed to roll back settings reset:", rollbackErrors);
    }
    throw error;
  }

  onSaved();
}
