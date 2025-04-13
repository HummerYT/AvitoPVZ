# AvitoPVZ

Для запуска проекта нужен установленный Docker, если его нет,
то можно установить `colima`

```
brew install colima
colima start
```

1. Чтобы запустить проект, используйте `make up` (чтобы завершить - `make down`)
2. Чтобы запустить юнит тесты, используйте `make unit`
3. Чтобы запустить e2e тесты, используйте `make integration`
4. Чтобы посмотреть покрытие тестами, используйте `make cover`

# Сложности реализации функционала

В процессе разработки функционала для работы с ПВЗ, приёмками и товарами возникли следующие основные сложности:

## 1. Динамическое формирование SQL-запросов

- **Фильтрация по датам:** При наличии необязательных параметров (startDate, endDate) запросы нужно было динамически
  дополнять условиями фильтрации, что требовало аккуратного формирования строки запроса и корректной передачи
  параметров.
- **Поддержание безопасности:** Динамическое формирование запросов подразумевает внимание к предотвращению SQL-инъекций,
  поэтому использовались параметры (placeholders) вместо прямой подстановки значений.

## 2. Объединение данных из нескольких таблиц

- **Вложенные данные:** Для получения полной информации о ПВЗ требовалось объединение данных из таблицы `pickup_point` с
  данными о приёмках (`receiving`), а также получение товаров (`goods`) для каждой приёмки.
- **Серия вложенных запросов:** Использование нескольких запросов (или объединённых запросов) для каждой сущности
  создаёт дополнительную сложность в обработке и агрегации данных, что приводит к необходимости явного обхода полученных
  результатов и их компоновки.

## 3. Реализация пагинации

- **Вычисление OFFSET:** Для корректной работы пагинации важно правильно вычислить смещение (OFFSET) и лимит (LIMIT) в
  SQL-запросах.
- **Согласованность с фильтрами:** При использовании дополнительных фильтров по дате пагинация должна работать корректно
  независимо от того, заданы ли фильтры или нет.

## 4. Валидация входных данных

- **Использование validator/v10:** Пакет [validator/v10](https://pkg.go.dev/github.com/go-playground/validator/v10)
  оказался полезным для валидации, но требовал правильного определения схем валидации, особенно для дат (RFC3339) и
  UUID.
- **Корректность данных:** Гарантия того, что все входные данные соответствуют ожидаемым форматам (например, дата-время
  или уникальные идентификаторы) до выполнения операций в базе.

## 5. Явное преобразование структуры данных для ответа

- **Сложная структура ответа:** Необходимо было аккуратно агрегировать данные из нескольких таблиц (ПВЗ, приёмки,
  товары) и формировать итоговый JSON-ответ с вложенными структурами.
- **Преобразование в fiber.Map:** Приведение итоговых данных к формату `fiber.Map` для явного и гибкого формирования
  ответа клиенту требовало дополнительного кода и внимания к уровню вложенности.

## 6. Обработка транзакций и блокировок

- **Атомарность операций:** Для операций, изменяющих данные (закрытие приёмки, удаление товаров), необходимо было
  обеспечить атомарное выполнение проверки и обновления данных.
- **Использование FOR UPDATE:** Для предотвращения гонок между конкурентными запросами применялась блокировка строк (
  `FOR UPDATE`) в рамках транзакций.
- **Уровень изоляции:** Выбор правильного уровня изоляции (например, `Serializable`) был критичен для сохранения
  целостности данных при параллельном доступе.

## 7. Написание тестов

Было недостаточно времени, чтобы покрыть сервис тестами