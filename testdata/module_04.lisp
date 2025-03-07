(module m1)
(defcolumns X)
;; Module without any column declarations to test alignment.
(module m2)
(defcolumns (X :i16@loob))
(defconstraint heartbeat () X)
