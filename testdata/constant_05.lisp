(defpurefun ((vanishes! :𝔽@loob) x) x)

(defconst ONE 1)
(defcolumns CT)
(defconstraint c1 ()
  (vanishes!
   (* (- CT (shift CT ONE)) (- (+ CT ONE) (shift CT ONE)))))
