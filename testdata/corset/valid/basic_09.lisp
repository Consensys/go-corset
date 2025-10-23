(defcolumns (CT :i4))

(defconstraint c1 () (== 0
                      ;; CT(i) == CT(i+1) || CT(i) + 1 == CT(i+1)
                      (* (- CT (shift CT 1)) (- (+ CT 1) (shift CT 1)))))

(defconstraint c2 () (∨
                      (== CT (shift CT 1)) (== (+ CT 1) (shift CT 1))))
