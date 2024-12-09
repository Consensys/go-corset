(defcolumns COUNTER)
(defalias CT2 CT1)
(defalias CT1 COUNTER)
(defconstraint heartbeat () CT2)
