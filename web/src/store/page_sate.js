
import { defineStore } from 'pinia';
import { LocalStieConfigUtils } from '@/util/localSiteConfig'


export const usePageState = defineStore('pageState', {
    // id: 'pageState',
    state: () => {
        return {
            isShowAddWayDialog: false,
            siteConfigData: LocalStieConfigUtils.getLocalConfig(),
            ShowDialogData: {}
        }
    },
    actions: {
        setShowAddWayDialog(status) {
            this.isShowAddWayDialog = status;
        },
        setSiteConfigData(configData) {
            this.siteConfigData = configData;
        },
    },
});
