(defcolumns (STAMP :i10))
;; STAMP[0] == 0
(defconstraint c1 (:domain {0}) (== 0 STAMP))
