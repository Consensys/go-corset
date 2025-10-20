(defcolumns (COUNTER :i16))
(defalias CT1 COUNTER)
(defalias CT2 CT1)
(defconstraint heartbeat () (== 0 CT2))
