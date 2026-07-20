import type { DefineComponent } from 'vue';
import type { DoraIconName } from '@/types/app';
import type { DashboardTone } from '@/types/business';

declare const CardNum: DefineComponent<{
  title: string;
  value: string | number;
  description?: string;
  badgeText?: string;
  iconName?: DoraIconName;
  routePath?: string;
  trendText?: string;
  trendType?: 'up' | 'down' | 'flat';
  tone?: DashboardTone;
  scopeLabel?: string;
  unit?: string;
}>;

export default CardNum;
