(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X)
;; intention is that shifts cancel.
(defconstraint c1 () (vanishes! (- X (shift (shift X -1) 1 ))))
(defconstraint c2 () (vanishes! (- (shift X 1) (shift (shift X -1) 2 ))))