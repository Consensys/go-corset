(defcolumns (COUNTER :i32))
(defalias CT COUNTER)
(defconstraint heartbeat () (== 0 CT))
