-- Очищаем таблицы перед заполнением (CASCADE удалит связанные данные)
TRUNCATE TABLE comments, user_favorites, movies RESTART IDENTITY CASCADE;

-- 1. Тестовые фильмы
INSERT INTO movies (id, title, original_title, description, release_date, duration_minutes, poster_url, created_at)
VALUES
    (1, 'Интерстеллар', 'Interstellar', 'Команда исследователей отправляется сквозь червоточину в поисках нового дома для человечества.', '2014-06-01', 169, 'https://images.example.com/interstellar.jpg', NOW() - INTERVAL '10 days'),
    (2, 'Начало', 'Inception', 'Вор, извлекающий секреты из подсознания во время сна, получает шанс вернуть свою прежнюю жизнь.', '2010-06-01', 148, 'https://images.example.com/inception.jpg', NOW() - INTERVAL '9 days'),
    (3, 'Темный рыцарь', 'The Dark Knight', 'Бэтмен с помощью полицмейстера и прокурора пытается очистить улики Готэма от преступности.', '2008-06-01', 152, 'https://images.example.com/dark_knight.jpg', NOW() - INTERVAL '8 days'),
    (4, 'Матрица', 'The Matrix', 'Хакер Нео узнает от загадочных повстанцев правду о виртуальной реальности.', '1999-06-01', 136, 'https://images.example.com/matrix.jpg', NOW() - INTERVAL '7 days'),
    (5, 'Побег из Шоушенка', 'The Shawshank Redemption', 'Бухгалтер Энди Дюфрейн обвинен в убийстве и приговорен к пожизненному заключению.', '1994-06-01', 142, 'https://images.example.com/shawshank.jpg', NOW() - INTERVAL '6 days');

-- Синхронизируем счетчик автоинкремента, чтобы новые INSERT не конфликтовали с явными ID 1..5
SELECT setval('movies_id_seq', (SELECT MAX(id) FROM movies));

-- 2. Избранное (user_id подставлены тестовые: 100, 101, 102)
INSERT INTO user_favorites (user_id, movie_id, created_at)
VALUES
    (100, 1, NOW() - INTERVAL '3 days'),
    (100, 2, NOW() - INTERVAL '2 days'),
    (100, 4, NOW() - INTERVAL '1 day'),
    (101, 1, NOW() - INTERVAL '4 days'),
    (102, 3, NOW() - INTERVAL '5 days');

-- 3. Комментарии к фильмам
INSERT INTO comments (movie_id, user_id, text, created_at, updated_at)
VALUES
    (1, 100, 'Гениальный фильм Нолана, музыка Ханса Циммера просто до мурашек!', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
    (1, 101, 'Пересматриваю уже в третий раз, каждый раз открываю что-то новое.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
    (1, 102, 'Финал каждый раз оставляет сильный эмоциальный след.', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours'),
    (2, 100, 'Концовка с волчком до сих пор не дает покоя...', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours'),
    (3, 101, 'Джокер в исполнении Хита Леджера — культовая роль.', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '10 hours');

-- Синхронизируем счетчик для комментариев
SELECT setval('comments_id_seq', (SELECT MAX(id) FROM comments));