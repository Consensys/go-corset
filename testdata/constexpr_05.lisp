(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y)
;; Y == 0
(defconstraint c1 () (if
                      (vanishes! (* 1 2))
                      (vanishes! X)
                      (vanishes! Y)))
