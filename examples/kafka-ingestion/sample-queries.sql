-- Sample SQL queries for the wikipedia-kafka datasource
-- Execute these in the Druid console at http://localhost:8888

-- 1. Basic data exploration - show recent ingested records
SELECT *
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR
ORDER BY __time DESC
LIMIT 10;

-- 2. Count records by language
SELECT 
  language,
  COUNT(*) as record_count,
  SUM(added) as total_added,
  SUM(deleted) as total_deleted
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR
GROUP BY language
ORDER BY record_count DESC;

-- 3. Activity by country over time
SELECT 
  TIME_FLOOR(__time, 'PT5M') as time_window,
  country,
  COUNT(*) as edits,
  SUM(delta) as net_change
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR
GROUP BY TIME_FLOOR(__time, 'PT5M'), country
ORDER BY time_window DESC, edits DESC;

-- 4. User activity analysis
SELECT 
  anonymous,
  robot,
  COUNT(*) as edit_count,
  AVG(added) as avg_added,
  AVG(deleted) as avg_deleted
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR
GROUP BY anonymous, robot
ORDER BY edit_count DESC;

-- 5. Real-time monitoring - recent activity by minute
SELECT 
  TIME_FLOOR(__time, 'PT1M') as minute,
  COUNT(*) as edits_per_minute,
  COUNT(DISTINCT page) as unique_pages,
  COUNT(DISTINCT user) as active_users
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '30' MINUTE
GROUP BY TIME_FLOOR(__time, 'PT1M')
ORDER BY minute DESC
LIMIT 30;

-- 6. Namespace analysis
SELECT 
  namespace,
  newPage,
  COUNT(*) as page_count,
  SUM(CASE WHEN delta > 0 THEN delta ELSE 0 END) as positive_changes,
  SUM(CASE WHEN delta < 0 THEN ABS(delta) ELSE 0 END) as negative_changes
FROM "wikipedia-kafka"
WHERE __time >= CURRENT_TIMESTAMP - INTERVAL '1' HOUR
GROUP BY namespace, newPage
ORDER BY page_count DESC;