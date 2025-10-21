(defcolumns (X :i16))
(module m1)
;; Module without any column declarations to test alignment.
(module m2)
(defcolumns (X :i16))
(defconstraint heartbeat () (== 0 X))
