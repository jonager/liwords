WITH duplicated_games AS
((SELECT
    created_at,
    player0_id AS player
FROM public.games)
UNION ALL
(SELECT
    created_at,
    player1_id AS player
FROM public.games)),
dau_report AS
(SELECT
	DATE_TRUNC('day',duplicated_games.created_at) AS day,
	COUNT(DISTINCT player) AS dau
FROM duplicated_games
LEFT JOIN public.users ON duplicated_games.player = users.id
WHERE NOT (users.internal_bot OR users.id IN (42,43,44,45,46))
GROUP BY 1),
bot_users AS
(SELECT
   id,
   internal_bot OR id IN (42,43,44,45,46) AS is_bot
FROM public.users),
pvp_games as
(SELECT
   created_at,
   games.player0_id,
   games.player1_id
 FROM public.games
 LEFT JOIN bot_users b1 ON games.player0_id=b1.id
 LEFT JOIN bot_users b2 ON games.player1_id=b2.id
 WHERE (NOT b1.is_bot)
   AND (NOT b2.is_bot)
 ),
duplicated_games_between_humans AS
((SELECT
    created_at,
    player0_id AS player
FROM pvp_games)
UNION ALL
(SELECT
    created_at,
    player1_id AS player
FROM pvp_games)),
dauh_report AS
(SELECT
    DATE_TRUNC('day',created_at) AS day,
	COUNT(DISTINCT player) AS dau_h
FROM duplicated_games_between_humans
GROUP BY 1)

SELECT
  dau_report.day,
  dau_report.dau,
  dauh_report.dau_h,
  TRUNC(100.0*dauh_report.dau_h/dau_report.dau,1) AS ratio
FROM dau_report
LEFT JOIN dauh_report ON dau_report.day = dauh_report.day
ORDER BY 1 DESC
