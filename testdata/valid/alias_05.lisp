(defcolumns (COUNTER :i12))
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () (== 0 CT2))
