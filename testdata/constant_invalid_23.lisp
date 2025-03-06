;;error:7:43-46:not permitted in const context
(defpurefun ((vanishes! :@loob) x) x)

(defconst (TWO :extern) 2)
(defcolumns X Y)
;; Y == X*X
(defconstraint c1 () (vanishes! (- Y (^ X TWO))))
