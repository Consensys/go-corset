(defpurefun ((vanishes! :@loob) x) x)

(defcolumns CT)
(defconstraint c1 () (vanishes!
                      (* (- CT (shift CT 1)) (- (+ CT 1) (shift CT 1)))))
