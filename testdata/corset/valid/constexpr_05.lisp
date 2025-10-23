(defcolumns (X :i16) (Y :i16))
;; Y == 0
(defconstraint c1 () (if
                      (== 0 (* 1 2))
                      (== 0 X)
                      (== 0 Y)))
