(defpurefun ((vanishes! :@loob) x) x)

(defextern TWO 2)
(defcolumns X Y)
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
