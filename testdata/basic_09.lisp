(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (CT :i4))
(defconstraint c1 () (vanishes!
                      ;; CT(i) == CT(i+1) || CT(i) + 1 == CT(i+1)
                      (* (- CT (shift CT 1)) (- (+ CT 1) (shift CT 1)))))
