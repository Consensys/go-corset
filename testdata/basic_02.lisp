(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :𝔽) (Y :𝔽))
(defconstraint c1 () (vanishes! (+ X Y)))
(defconstraint c2 () (vanishes! (+ Y X)))
