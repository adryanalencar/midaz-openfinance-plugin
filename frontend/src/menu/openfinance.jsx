const openfinanceMenu = {
  id: 'group-openfinance',
  title: 'Open Finance Sync',
  type: 'group',
  children: [
    {
      id: 'openfinance-dashboard',
      title: 'Sync Overview',
      type: 'item',
      url: '/openfinance',
      icon: 'IconBuildingBank'
    },
    {
      id: 'connected-accounts',
      title: 'Bank Accounts',
      type: 'item',
      url: '/openfinance/accounts',
      icon: 'IconWallet'
    },
    {
      id: 'consents-management',
      title: 'Consents Lifecycle',
      type: 'item',
      url: '/openfinance/consents',
      icon: 'IconShieldCheck'
    },
    {
      id: 'statement-transactions',
      title: 'Statement Items',
      type: 'item',
      url: '/openfinance/transactions',
      icon: 'IconArrowsExchange'
    }
  ]
};

export default openfinanceMenu;
