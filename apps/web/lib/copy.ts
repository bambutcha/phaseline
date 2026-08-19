export const layoutRu: Record<string, string> = {
  mixed: "смешанный реголит",
  solar_belt: "солнечный пояс",
  ridge_wall: "хребет",
  crater_field: "кратерное поле",
  dust_sea: "пылевое море",
  cold_front: "холодный фронт",
};

export const crisisRu: Record<string, string> = {
  dust_storm: "Кризис: пылевая буря. Пыльные клетки ещё дороже.",
  solar_flare: "Кризис: вспышка. Солнечные клетки больше не заряжают.",
  cave_in: "Кризис: обвал. Один кратер закрыт навсегда.",
  vip_override: "Кризис: VIP. Появился срочный заказ — реши, брать ли.",
  comm_blackout: "Кризис: радиотишина. Батарея садится быстрее.",
};

export const terrainRu: Record<string, string> = {
  base: "база — стой и заряжайся",
  regolith: "реголит — обычная клетка",
  solar_plateau: "солнце — заряд в движении",
  crater: "кратер — медленно",
  ridge: "хребет — жрёт батарею",
  dust_field: "пыль — чуть медленнее",
  cold_sink: "холод — дороже клетка",
};

export const hexColors: Record<string, string> = {
  base: "#d4b44a",
  regolith: "#8a9098",
  solar_plateau: "#e0c05a",
  crater: "#4a322c",
  ridge: "#6d6360",
  dust_field: "#b49a70",
  cold_sink: "#2e86ab",
};

export const statusRu: Record<string, string> = {
  queued: "ждёт тебя",
  accepted: "принят",
  in_transit: "в кузове",
  delivered: "сдан",
  failed: "срыв",
  expired: "срок",
  lost_to_shadow: "тень забрала",
};

export const weightRu = { light: "лёгкий", medium: "средний", heavy: "тяжёлый" } as const;
export const riskRu = { low: "низкий", medium: "средний", high: "высокий" } as const;
export const urgRu = { low: "спокойно", medium: "средняя", high: "горит" } as const;

export const rejectRu: Record<string, string> = {
  swift_no_heavy: "Это тяжёлое. Нажми Hauler сверху, потом заказ снова.",
  slots_full: "У ровера оба слота заняты. Дождись сдачи.",
  battery: "Батареи не хватит. Съезди на золотую базу или солнечную клетку.",
  shadow: "В тень нельзя. Это стена — клетка уже мертва.",
  no_path: "Пути нет. Тень перекрыла клетку или кратер обвалился.",
  wrong_rover: "Этот заказ уже везёт другой ровер. Нажми его сверху.",
  stranded: "Этот ровер сел. Переключись на второго.",
};

export const outcomeTitle: Record<string, string> = {
  colony_saved: "Колония спасена",
  pyrrhic: "Пиррова победа",
  signal_lost: "Сигнал потерян",
};

export const endReasonRu: Record<string, string> = {
  time: "Смена закрыта по таймеру — тень ещё могла быть далеко. Это нормально: цель считается в конце.",
  shadow: "Карта ушла в тень.",
  stranded: "Оба ровера сели.",
  delivered: "Все заказы сданы.",
  cleared: "Больше нечего везти — смена закрыта.",
};

export const tutorialSlides = [
  {
    title: "Линия наступает.",
    body: "Золотая полоса — терминатор. Клетки уходят в тень: батарея садится быстрее, заказы сгорают. В тёмную клетку войти нельзя.",
  },
  {
    title: "Два ровера. Всех не спасти.",
    body: "Swift быстрый, но не везёт тяжёлое. Hauler медленнее и тащит стержни и гелий-3. У каждого 2 слота и своя батарея.",
  },
  {
    title: "Клик — едет. Новый клик плавно сворачивает.",
    body: "На ходу клик по другой клетке: ровер остаётся на текущей позиции, доезжает клетку и сворачивает. Без отката на базу.",
  },
  {
    title: "Цель смены.",
    body: "Набери 100 очков колонии за 3 минуты. Роверы быстрые, тень медленнее. Двух заказов мало. Бирюзовые кассеты — ещё немного очков, если заказы кончились. Если набрал раньше — смена не обрывается.",
  },
];

export const legend = [
  { color: "#d4b44a", label: "база — зарядка, если стоишь" },
  { color: "#e0c05a", label: "солнце — заряд даже в движении" },
  { color: "#8a9098", label: "реголит — обычная клетка" },
  { color: "#6d6360", label: "хребет — жрёт батарею" },
  { color: "#4a322c", label: "кратер — медленно" },
  { color: "#2e86ab", label: "холод — дороже клетка" },
  { color: "#b49a70", label: "пыль — чуть медленнее" },
  { color: "#5ec8c8", label: "кассета — +8 колонии, без сдачи" },
];
