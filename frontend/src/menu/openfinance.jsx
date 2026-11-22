const openfinanceMenu = {
  id: 'group-openfinance',
  title: 'Open Finance',
  type: 'group',
  children: [
    {
      id: 'openfinance-dashboard',
      title: 'Visão Geral',
      type: 'item',
      url: '/openfinance',
      icon: 'IconBuildingBank'
    },
    {
      id: 'connected-accounts',
      title: 'Contas Bancárias',
      type: 'item',
      url: '/openfinance/accounts',
      icon: 'IconWallet'
    },
    {
      id: 'consents-management',
      title: 'Autorizações',
      type: 'item',
      url: '/openfinance/consents',
      icon: 'IconShieldCheck'
    },
    {
      id: 'statement-transactions',
      title: 'Extrato Consolidado',
      type: 'item',
      url: '/openfinance/transactions',
      icon: 'IconArrowsExchange'
    }
  ]
};

export default openfinanceMenu;
