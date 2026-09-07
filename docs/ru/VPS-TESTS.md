# Тесты VPS (русский)

**EN:** [../VPS-TESTS.md](../VPS-TESTS.md)

```bash
sudo freshvps-tests          # меню
sudo freshvps-tests --default
sudo bash install.sh --tests
```

В конце — сводная таблица OK/FAIL. Рабочий каталог `/tmp/freshvps-tests.*` удаляется при выходе. Пакеты, которые поставили чужие скрипты (sysbench и т.д.), не откатываем.
