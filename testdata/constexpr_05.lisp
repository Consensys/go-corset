(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16))
;; Y == 0
(defconstraint c1 () (if
                      (vanishes! (* 1 2))
                      (vanishes! X)
                      (vanishes! Y)))
