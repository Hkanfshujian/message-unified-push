export interface PageState {
  isShowAddWayDialog: boolean;
  siteConfigData: any;
  ShowDialogData: any;
}

export interface PageStateActions {
  setShowAddWayDialog(status: boolean): void;
  setSiteConfigData(configData: any): void;
}

export declare const usePageState: () => PageState & PageStateActions;