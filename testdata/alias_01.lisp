(defpurefun (vanishes! x) (== 0 x))

(defcolumns (COUNTER :i32))
(defalias CT COUNTER)
(defconstraint heartbeat () (vanishes! CT))
